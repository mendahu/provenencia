# Review — S10-39 Preferred portrait

**Plan:** [`S10-39-preferred-portrait.md`](S10-39-preferred-portrait.md)
**Lint:** `python3 scripts/check-plan-docs.py` clean before round 1 (except this file's link).

## Round 1 (2026-10-17)

Maintainability: **Fits with changes**. Executability: **Not ready** (X-1).

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | `canonicalentities` importing `likenesses` makes an internal `likenesses` test that reaches `promote` an import cycle | **Fixed.** D-14: `likenesses_test.go` is `package likenesses_test`; STEP-2 changes the package line if S10-35 left it internal; new stop condition if its tests use unexported names. `go list -deps` shows no production cycle |
| X-2 | major | STEP-3 says `writes.Call`, T-2 needs `writes.Run` | **Fixed.** D-14 and STEP-3: T-1 setup `writes.Call`, T-2 `writes.Run` |
| X-3 | major | Nothing fills `LikenessDetail.preferred` in the handler | **Fixed.** D-7: S10-35's `attachLikenesses` sets `Preferred: l.Preferred` |
| X-4 | minor | `call_test.go` has two `101`s | **Fixed.** D-6 names both |
| X-5 | minor | T-5 / T-2 fixture values and the name Observation helper unspecified | **Fixed.** Shades 100 / 200 (T-5), 150 (T-2); invalid cases are another Person's likeness and a fresh UUIDv7, so no name helper is needed. AC-3.3 reworded to match |
| X-6 | minor | pbxproj ids; `evaluate-ui-component` missing from Skills to load | **Fixed.** D-9 names `A90A…05` / `A90A…06`; skill added to the header |
| X-7 | minor | Review log missing | **Fixed.** This file |
| X-8 | info | protoc, Xcode, S10-35 code not available here | Expected; the Machines line and status `Blocked — on S10-35` cover it |
| M-1 | major | Fourth hand-rolled menu keyboard reset; the kit should own it | **Fixed (user, 2026-10-17: consolidate in this PR).** D-16 and STEP-8: `PVContextMenuKeyboard.opened(titles:)` and `pvContextMenu(…, keyboardTitles:)`; the Source cover and composer clear menus drop their resets; `PVSelect` keeps its seeded highlight. T-8 and §10 step 6 |
| M-2 | major | "Which likeness does this Observation name" decided in two places | **Fixed.** D-2 adds `likenesses.ByObservation`; `ForEntities` and D-3 both call it |
| M-3 | minor | Write tested in another package to avoid a fixture copy | **Declined for this PR.** The shared fixture touches S10-35's tests and other packages; Out line and follow-up task. D-14 keeps the tests beside the fixture |
| M-4 | minor | Fifth private `nullBlob` | **Declined for this PR.** Out line and follow-up task for `database.NullBlob` |
| M-5 | minor | Handle invalidation reloads Events and Places lists | **Logged.** Existing registry behavior; no new mechanism here |
| nit | nit | `VocabularyToast` is the app-wide toast under a vocabulary name | Logged |

## Round 2 (2026-10-17)

Fresh executability reviewer on the plan and this log. Every round-1 answer holds. **Ready with changes**, no blocker.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-9 | minor | Where the pointer sits between scan and grouping; what to do with a `ForEntities` error in D-3 | **Fixed.** D-2: per-entity local map; D-3: error returned as is |
| X-10 | minor | Success toast form unstated | **Fixed.** D-9 names `VocabularyToast(…, tone: .success)` |
| X-11 | nit | `observations.Delete` route in T-1 | **Fixed.** T-1 names `writes.Call` + `observations.Delete(tx, userID, id)` |

## Verdict

Maintainability **Fits with changes** (applied). Executability **Ready with changes** (applied). No open blocker. M-1 was answered by the user: consolidated in this PR (D-16). Status stays `Blocked — on S10-35 landing` until §4's (S10-35) rows are redone against its Landed note.
