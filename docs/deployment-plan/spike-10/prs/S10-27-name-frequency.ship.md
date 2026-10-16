# Ship log — S10-27 Name frequency in this project

**Plan:** [`S10-27-name-frequency.md`](S10-27-name-frequency.md) · **Branch:** `s10-27-name-frequency` · **PR:** pending
**State:** fixing 1

## Round 0 — implement
Head `67f9df9d`. Gate PASS (orchestrator re-ran). PR #409 (draft).
Report deviations: three existing `core/database/matching` tests (`TestForSubjectPersons`, `TestForSubjectComparesNameParts`, `TestForEntityMergeHints`) had absolute expected scores lowered because `rank` counts loaded candidates (D-3) and fixtures share surnames; implementer hand-checked arithmetic. Also `core/graphalign/types.go` imports `match` for `Stats.Words`; `matching.go` gained a `nameProperty` var. T-8 uses a new handle commit; T-11 asserts query count/undiscounted place score.

## Round 1 — review
| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Every handle commit forces a full Person-name rescan for the next Promote | Follow-up (to user to acknowledge) | Fix (patch `Words` on commit) reopens D-3/D-10 ("no new graph slot", counts live and die with `Stats`). Plan Risk row 1 accepted the rebuild; measured about 70 ms at 20,000 Persons against the 250 ms DoD bar. `ValueFreq`, fan-out and kinship already rebuild the same way. Preparatory PR: make Promote stats patchable on commit |
| MNT-1 | MNT | major | `rank` counts the candidate set, which the planned blocking change narrows | Fix now → F-1 (comment and doc lines only); shared count source → Follow-up | verified: `matching.go` `loadCandidates` comment says "narrow it here" and `docs/matching.md` blocking bullet says "Scoring does not change". Moving the count to one shared source reopens D-3 |
| MNT-2 | MNT | minor | `rank` and `WithNameComparer` pick name Features differently | Fix now → F-1 | `matching.go` is already touched by F-1; select by comparer type, drop `nameProperty` |
| MNT-3 | MNT | minor | D-11 note sits on the `ProfiledKinds` comment, not `DefaultProfile` | Fix now → F-2 | D-11 says the `DefaultProfile` comment states it; comment move only |
| MNT-4 | MNT | nit | `ScoreSimilarity` has two adjacent float64 parameters | Logged | Revisit when S10-28 adds a term; pass a struct then |
| CON-1 | CON | minor | T-11 cannot fail if the name-Feature check is deleted | Follow-up | needs a test seam (`export_test.go` hook); the Landed note already admits `Frequency` is not visible to tests |
| PRF-2 | PRF | minor | `CountWords` allocates a map per handle | Follow-up | `frequency.go` is not in the fix plan; constant factor only |
| UX-1 | UX | minor | Promote shows different weights for two identical Agree rows with no reason | Follow-up | needs `Discount` passed through FFI and new copy; the plan puts both out of scope |
| SEC | SEC | — | No findings | — | — |

### Fix plan 1
**F-1 — Name the coupling to the blocking seam, and pick name Features by comparer type** (from MNT-1, MNT-2)
Files: `core/database/matching/matching.go`, `docs/matching.md`.
Failing test first: none. Comments, docs and a refactor with no behavior change; the existing matching tests are the guard.
Do:
- In the `loadCandidates` comment and the `docs/matching.md` blocking bullet, say that `rank` takes its name-word counts from this set. Blocking must therefore count a separate, unnarrowed population, and "Scoring does not change" holds only if it does.
- In `rank`, find name Features the same way `Profile.WithNameComparer` does (`Comparer.(match.NameComparer)`). Collect each candidate's names from those Features' Properties in a small named helper next to `rank` (for example `candidateNames`). Remove `nameProperty`.
Check: `go test -tags fts5 ./core/database/matching` passes unchanged; `scripts/check-pr.sh` PASS.

**F-2 — Put the D-11 paragraph on `DefaultProfile`** (from MNT-3)
Files: `core/match/registry.go`.
Failing test first: none (comment move).
Do: move the "DefaultProfile is the shipped algorithm…" paragraph, including the sentence about the `MinScore` cut seeing the discounted score, above `func DefaultProfile`. Leave the one `ProfiledKinds` line above `ProfiledKinds`. Rewrap to the file's width.
Check: `go doc ./core/match DefaultProfile` shows the paragraph; `scripts/check-pr.sh` PASS.

