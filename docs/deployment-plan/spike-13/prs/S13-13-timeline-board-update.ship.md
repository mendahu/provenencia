# Ship log — S13-13 Timeline section: the board's 2026-11-03 update

**Plan:** [`S13-13-timeline-board-update.md`](S13-13-timeline-board-update.md) · **Branch:** `s13-13-timeline-section` · **PR:** #691
**State:** implementing

Base: branch @ `905d2083` (PR #691, ready for review, CI green on `d87045a1`), on `main` @ `92f87dba`. The user changed the board on 2026-11-03 and asked for the update on the same branch; plan Planned after two review rounds (user answers: age in Go, `.compact` reshaped, 29 February → 1 March). The first ship log, [`S13-13-timeline-section.ship.md`](S13-13-timeline-section.ship.md), covers the section as first built.

## Round 0 — implement
Head `ef3d0259`. Gate PASS (orchestrator run against `92f87dba`: all nine checks). Report deviations: (1) `PersonTimelinePlacement.lifeline` is optional, nil until every item has a height; (2) the flag font is `PVTypeScale.micro` (11); (3) `Row.age`, `Row.ageHelp` and one `ageText(_:)` rule; (4) `PersonTimelineAgeFlag(isDashed:)` from "undated row"; (5) the curve ends at the flag's leading edge (x = 96) with no end dot; (6) docs as STEP-7; (7) nothing outside the plan. Not run in the app: click-to-scroll over `PersonTimelineStack`, the frames, tethers on scroll, Wide resize.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, strings. SEC+PRF: `core/datevalue` and `core/valuecompare` (matching's leap-day windows via `ShiftYears`, `core/subjectprofile`), `api/ffi/**`, `core/database/**`, feature models and list layout.
