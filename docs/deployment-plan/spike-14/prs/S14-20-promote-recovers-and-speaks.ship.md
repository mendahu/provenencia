# Ship log — S14-20 The Promote page recovers and speaks

**Plan:** [`S14-20-promote-recovers-and-speaks.md`](S14-20-promote-recovers-and-speaks.md) · **Branch:** `s14-20-promote-recovers-and-speaks` · **PR:** #706
**State:** ready

## Round 0 — implement
Head `9b029f3b`. Gate PASS (orchestrator). Report deviations: hold-based tests enable `proposePromoteGraphAlignmentHold` only after `load()`; `mappingTheRestWhileProposingSendsTheWholePage` clears the hold before `release()` so one release matches the plan while the coalesced follow-up can finish. Manual §10 not run.

## Round 1 — review
Reviewers: CON (Composer), MNT, UX, SEC+PRF. Skipped: none. UX for `macos/App/**` and `Localizable.xcstrings`; SEC for `core/graphalign`; PRF for `core/database/promotealign` and `Features/Promote` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Manual §10 not run | Rejected (Needs you at finish) | Ship-pr finish assigns §10 to the user; automated AC evidence is T-8 / T-7. Not an implementer code fix |
| MNT-1 | MNT | major | Scripted propose error is read off the store lock | Fix now → F-1 | verified: `FakeStore.swift:1481` throws outside `withState`; `updateArtifact:767` and `applyPromoteGraphAlignment:1560` throw inside |
| MNT-2 | MNT | minor | Scoring lead still says provenance scales | Fix now → F-2 | verified: `promote-graph-alignment.md:137` opens with "Provenance scales…" then says it does not |
| UX-1 | UX | major | A later proposal failure stays off-screen and silent | Fix now → F-5 | user (2026-11-04): Fix. verified: clear `proposeError` on new attempt; update T-8 (supersedes “not on a second proposal error while the first still shows”) |
| UX-2 | UX | minor | Try again loses keyboard focus when the same failure remounts | Fix now → F-3 | verified: `.id(model.loadAttempt)` remounts the whole `PVRetryCallout` (D-8); focus is not restored. Remount stays; restore focus on `promote.tryAgain` after a failed attempt |
| UX-3 | UX | minor | Read again on a cached graph does not speak a second failure | Fix now → F-4 | verified: Evidence graph `PVRetryCallout` has no attempt id; `isRereadingGraph` keeps the callout mounted with unchanged text. File this PR changed |
| SEC / PRF | SEC+PRF | — | — | — | Clean: no findings |

Nits (logged): `CON-2 · ship log / PR body · hold-timing deviations already in Landed note` · `MNT-3 · FakeStore.swift:1547 · apply comment still says refuses only stale revision`

### Fix plan 1

**F-1 — Throw propose error inside `withState`** (from MNT-1)
Files: `macos/App/Platform/FakeStore.swift`.
Do: after `await proposePromoteGraphAlignmentHold.wait()`, throw `proposePromoteGraphAlignmentError` from inside `withState`, matching `updateArtifact` after its hold. Do not snapshot the error before the wait.
Check: throw inside `withState`; scripts/check-pr.sh PASS.

**F-2 — Lead the scoring paragraph with "not scaled by provenance"** (from MNT-2)
Files: `docs/promote-graph-alignment.md`.
Do: rewrite the §6 bold lead so it states graph alignment does not scale by provenance yet, then name credibility and transcription certainty as the inputs that would. Match §12.
Check: paragraph no longer opens with "Provenance scales each comparison"; scripts/check-pr.sh PASS.

**F-3 — Restore focus on Try again after a remounted load failure** (from UX-2)
Files: `macos/App/Features/Promote/PromoteView.swift`, `PVRetryCallout` if needed.
Do: keep D-8's `.id(model.loadAttempt)` remount. After a failed load/retry settles, move keyboard focus to Try again (`promote.tryAgain`).
Check: T-7 still passes; scripts/check-pr.sh PASS.

**F-4 — Remount Evidence graph read-error callout on a repeated failure** (from UX-3)
Files: `EvidenceGraphView.swift`, `EvidenceGraphModel.swift`.
Do: bump an attempt id when a reread ends still failed; `.id(...)` on the `PVRetryCallout`.
Check: callout remounts on a second identical failure; scripts/check-pr.sh PASS.

**F-5 — Clear `proposeError` when a new proposal attempt starts** (from UX-1; user 2026-11-04)
Files: `PromoteModel.swift`, `PromoteModelTests.swift`, plan Landed note.
Do: clear `proposeError` in `proposeAfterDecision()` and at the start of `done()`; update T-8 so a second decision's failure bumps `noticeToken`. Landed note: T-8 clause superseded by user.
Check: `aNewErrorBumpsTheNoticeToken` passes with the second bump; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `6733b5ce`. Gate PASS (orchestrator). F-1 – F-5 done. Deviations: F-4 bumps `graphReadAttempt` in the model after seeing `isFetching` (not view `onChange`); F-3 adds optional `retryFocused` on `PVRetryCallout`.

Round-1 minors resolved by orchestrator read: MNT-2 (`§6` lead is “does not scale by provenance yet”); UX-2 (`retryFocused` binding on Promote load callout); UX-3 (`graphReadAttempt` + `.id` on Evidence graph callout).

## Round 2 — review
Reviewers: MNT, UX (each had a Fix-now major). Not sent: CON (CON-1 rejected), SEC+PRF (clean). Scope: `9b029f3b...6733b5ce`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) | Resolved | throw inside `withState` after hold (`FakeStore.swift:1480-1486`) |
| UX-1 | UX | major | (round 1) | Resolved | `proposeError = nil` in `proposeAfterDecision` / `done`; T-8 expects second bump |
| MNT-2, UX-2, UX-3 | — | minor | (round 1) | Resolved | orchestrator read in Round 1 — fix |
| MNT-4 | MNT | blocker | Graph reread waits by polling the handle | Fix now → F-6 | verified: `noteGraphReadFailureWhenSettled` polls 2s / 10ms; `prepare` already uses `readyValue` |

Nits (logged): —

### Fix plan 2

**F-6 — Await `readyValue` after Evidence graph reread** (from MNT-4)
Files: `EvidenceGraphModel.swift`, tests if needed.
Do: after `session.reread(graphKey)`, `await session.readyValue(graphKey)`, then `graphReadFailedSettled()`. Delete the poll. Bump gated on `graphReadError != nil`.
Check: no poll in `EvidenceGraphModel`; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `11f9761b`. Gate PASS (orchestrator). F-6 done: `readGraphAgain` awaits `readyValue` then `graphReadFailedSettled()`; `rg ContinuousClock|Task.sleep|noteGraphReadFailureWhenSettled` empty. Merged `origin/main` @ `330243ad` → head `960520ce`; gate PASS again.

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-3 (apply comment nit) · to user: Manual §10 · CON miss: none
