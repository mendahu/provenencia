# Ship log — S11-13 Migrate Promote scoring onto subject compare

**Plan:** [`S11-13-promote-subject-compare.md`](S11-13-promote-subject-compare.md) · **Branch:** `s11-13-promote-subject-compare` · **PR:** #487
**State:** ready

## Round 0 — implement
Base `main` @ `6ecfd88f`. Head `3539b77f`. Gate PASS (orchestrator run). Report: no stop. Deviations are the plan §1 listed ported-test changes (b)–(d), recorded in the plan's Landed note: `TestSeedFromNewlyFixedIntoUnpromoted` passes `Stats{FanOut: 1}`; `TestNewEventsDisagreeWhenTheirParticipantsDo` dates its events; `TestPartialNamesAreWeak` pair fixed for the sex-mismatch score; `TestRowReasons` weak case and surname-only case changed; `TestRankLoserDoesNotNominate` → `TestCandidateLoserDoesNotNominate`; `TestProposeOverlappingBoundShowsTheDateAndResembles` uses a residence (two births "before" different dates conflict at -20); `TestProposeJudgesAgainstTheMajority` caches the minority name as `outvoted`; `layerValues` skips Observations carrying a value Subject; test helpers `datedEvent`, `metasOf`, `mergeMetas`. STEP-2/3/4 one commit (D-17). Manual check not run (no app session).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (matching layers, `api/ffi/handlers`, `core/database/**`). Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or new error codes; the FFI diff adds none).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker → major | Seeding profiles every catalog person and every event they take part in, on each repropose | Follow-up (user, 2026-10-18) → S11-15; F-3 adds the note | Reviewer benchmark (`bulkCanon`, 20k persons / 30k events): warm repropose 40 ms → 160 ms, 49 MB → 181 MB; cold 317 → 605 queries. The code does what D-10 prescribes ("before seeding a kind it calls `profile` on that kind's `g.Kind` nodes"), so a fix changes D-10. Shortlist (S11-15) and memoization (S11-21) are planned Out lines |
| MNT-1 | MNT | major | `add-match-rule` does not teach `subjectcompare.Points`, `Rules.Lead`, `Rules.Derived` | Fix now → F-1 | verified: SKILL.md:30 describes the points rule in prose and names none of the three; callers already exist in `exhibits.go` and `align.go` |
| MNT-2 | MNT | minor | `NormalizeForm` doc names the deleted `core/match` | Fix now → F-2 | verified: `core/valuecompare/words.go:18`; a stale reference this PR's deletion created |
| MNT-3 | MNT | minor | Similarity read out of a `propertycompare.Result` in two places beside `Points` | Follow-up | Contradicts D-12's `Points(outcome, similarity, w)` signature; S11-19 adds the third caller and can reshape `Points` then |
| MNT-4 | MNT | minor | `loadExhibitObservations` builds `autoreconcile.Value` beside `decode` | Follow-up | Plan Out line: S11-19 / S11-25 replace that read; name `decode` in their Codebase fit |
| MNT-5 | MNT | minor | Pin-test helper `obs` copies production's `exhibitObs` construction | Follow-up | Fix plan does not otherwise touch `exhibits.go`; goes with S11-19 / S11-25's rewrite of `pairExhibits` |
| PRF-2 | PRF | minor | `pairExhibits` compares every member before checking `Compatible` | Follow-up | Bounded by one row's members; `exhibits.go` not otherwise in the fix plan; S11-19 / S11-25 replace the loop |

Nits (logged): `CON-1 · promotealign_test.go · fixture changes beyond D-16 (a)–(d), disclosed in the Landed note` · `MNT-6 · core/graphalign/align.go:66 · rename kindCands → kindHandles` · `MNT-7 · core/graphalign/registry.go:9-12 · header names only Config.Value` · `MNT-8 · promotealign/layer.go:77 · one build-order rule (lifespan kinds last) on both sides` · `MNT-9 · promotealign/profiles.go:202,259 · g.Follow called twice per person (also PRF outside-lens)` · `MNT-10 · core/matchregistry/value.go:16-82 · comments cite deleted match.* symbols` · `MNT-11 · layer.go / profiles.go · propertiesByID built twice per proposal`

PRF-1 to user; answered 2026-10-18: "defer to S11-15 as a major follow-up. Add a note to S11-15's scope that life events load only for shortlisted handles."

### Fix plan 1
**F-1 — Teach the layer-3 helpers in `add-match-rule`** (from MNT-1)
Files: `.agents/skills/add-match-rule/SKILL.md`. Skill: `add-match-rule`.
Failing test first: none (docs).
Do: in the Layer 3 bullet name `subjectcompare.Points` (what one property adds to evidence, for the score and the sheet), `Rules.Lead` (maps a pooled or period property to its lead) and `Rules.Derived` (a lead with no catalog row, the lifespan; a new derived property is declared in the registry, not detected from a missing meta). Add a "Do not": write a points switch or a lead lookup outside `core/subjectcompare`.
Check: scripts/check-pr.sh PASS.

**F-2 — Drop the stale `core/match` reference** (from MNT-2)
Files: `core/valuecompare/words.go`.
Do: `NormalizeForm`'s doc names its current users (the name sort key and text value compare), not `core/match`.
Check: `rg -n 'core/match\b' core --type go` prints nothing; scripts/check-pr.sh PASS.

**F-3 — Note PRF-1 on S11-15's scope entry** (from PRF-1, user call 2026-10-18)
Files: `docs/deployment-plan/spike-11/deployment-plan.md` (`#### S11-15 — Shortlist index`).
Do: add one sentence to **Ships.**: when the shortlist replaces the kind scan in `graphalign.Candidates`, the Promote profiler (`core/database/promotealign/profiles.go`, `profiler.profile`) loads and builds life events only for shortlisted handles (and the fixed and BFS-reached ones), not for every handle of the kind. S11-13 profiles every person and their events per repropose (PRF-1 in its ship log: warm repropose 40 → 160 ms on the 20k-person fixture). Leave the map row's status unchanged.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/deployment-plan.md` 0 errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `90c06785`. Gate PASS (orchestrator run). F-1 done (`1d8a3b51`); F-2 done (`1040250a`, plus `90c06785` for the same stale phrase in `words_test.go`, needed for F-2's check); F-3 done (`922e5612`). MNT-2 **Resolved** (read by orchestrator). PRF-1 note **Resolved** on S11-15's Ships (read by orchestrator).

## Round 2 — review
Reviewers: MNT (MNT-1 was a Fix-now major), scope `3539b77f...90c06785`. Not sent: CON (clean), SEC+PRF (PRF-1 deferred by the user, no Fix-now blocker or major).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) skill does not teach the layer-3 helpers | **Resolved** | `1d8a3b51`: Layer 3 bullet names `Points`, `Rules.Lead`, `Rules.Derived`; Do-not added; matches `subjectcompare.go:133,148,161` |

Nits (logged): `MNT-12 · .agents/skills/add-match-rule/SKILL.md:30 · Rules.Derived is also true for a lifespan's Span; say "a lifespan's lead or Span"`

No new blockers; no Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: PRF-1 (major, user call → S11-15, noted on its Ships), MNT-3, MNT-4, MNT-5, PRF-2 (S11-19 / S11-25) · to user: PRF-1 (answered: defer) · CON miss: none
