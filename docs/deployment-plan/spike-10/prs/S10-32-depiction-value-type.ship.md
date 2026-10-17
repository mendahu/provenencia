# Ship log — S10-32 Depiction value and `likeness` on Person

**Plan:** [`S10-32-depiction-value-type.md`](S10-32-depiction-value-type.md) · **Branch:** `s10-32-depiction-value` · **PR:** #430
**State:** fixing 1

The plan refresh (`36ef0065`) is the branch's first commit, on the user's call (2026-10-17): no separate docs PR. Base `ae26c629`.

## Round 0 — implement
Head `0532a08b`. Gate PASS (orchestrator run, against `ae26c629`). Report deviations: STEP-6's `go list -deps ./core/database/...` check was wrong in the plan (`sources` and `sourcemetadata` already import `core/writes` on `main`); `go list -deps ./core/database/observations` shows neither `ingest` nor `writes`. Depiction insert body is `resolveDepiction` in `depiction.go`; `releaseReplacedValue` runs from one table loop in `updateOne`; STEP-6 and STEP-7 landed together; extra T-6/T-7 rows and test helpers. Mac manual check not run.

## Round 1 — review
Five lenses on `ae26c629...0532a08b`. Verdicts: CON changes needed; MNT, UX, SEC, PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | T-7's date-release clause has no test | Fix now → F-1 | Plain: no test replaces a date value on update (`rg 'FROM date_values'` hits only the delete row); reviewer ran it via `-overlay`, behavior correct |
| CON-2, MNT-2 | CON, MNT | minor | Whitespace-only re-indent of `properties.Create` return block | Fix now → F-2 | `git diff -w` drops the hunk. Not in a file the fix plan touches, but it reverts this PR's own noise; zero risk |
| CON-3 | CON | minor | Manual check §10.1 not run | Follow-up → user's manual check | The user runs the manual check before merge (ship-pr step 8) |
| MNT-1 | MNT | minor | `ValidateDepictionRegion` decodes twice via a second decoder | Follow-up | `core/locator` split into `decode` + `validateDoc`; pairs with SEC-1's point cap, before S10-54 |
| MNT-3 | MNT | minor | `resolvedValue` carries value changes three ways | Follow-up | Contradicts D-8's named fields; log with M-8 (declared value-column list) |
| MNT-4 | MNT | nit | `sameStrings` duplicates `slices.Equal` | Logged | |
| MNT-5 | MNT | nit | Region trimmed twice; comment names `Insert` | Logged | |
| UX-1 | UX | minor | VoiceOver reads Likeness value type as *Text* | Follow-up → S10-54 | Plan Risks row accepts it; S10-54 adds the label |
| UX-2 | UX | nit | "More than one is allowed." vs "More than one value is allowed." | Logged | D-4 fixes the text |
| SEC-1 | SEC | minor | No cap on region point count; O(n²) self-intersection | Follow-up → before S10-54 | Same gap exists for Citation locators; unreachable until S10-54 |
| SEC-2 | SEC | minor | `StageDepictions` holds every staged PNG; no per-call cap | Follow-up → S10-54 | Handler-level cap when the FFI field lands |
| SEC-3 | SEC | minor | `Listed.Depiction.RegionJSON` unvalidated on read | Follow-up → S10-34, S10-36 | Validate and clamp before drawing or cropping |
| PRF-O1 | PRF→CON | — | `releaseSnapItem` deletes an owned row without counting | Rejected | Unique partial index `observations_value_*_id_uidx` (`000030.sql`, `000052.sql`) guarantees one owner; D-9 |
| SEC-O1 | SEC→CON | — | Replaced image File gets no `file` delete audit change | Rejected | D-9: "a released File has no audit change, as on delete today"; broader fix logged as M-R3 in the review log |

### Fix plan 1

**F-1 — Test a date value released by an update** (from CON-1)
Files: `core/database/observations/observations_test.go`.
Failing test first: `TestObservations/date_replaced_by_another_Property_releases_its_value` — on an Event, save a `date` Observation (point, `StartYear` 1844) with `saveCitation`; look up `eventNameProp := properties.Lookup(c, "event_name", properties.OriginProvenencia)`; update the Observation to `eventNameProp` with `ValueText: "Birth", HasText: true` via `updateObservation`.
Do: assert `rowCount(SELECT COUNT(*) FROM date_values WHERE id = ?, oldDateID) == 0` and `changeKinds(changes, "date_value")` is exactly `["date_value/delete"]`. Reuse the existing helpers; no product code change. It passes on the current code (behavior is correct); confirm it fails if the `observations.value_date_id` entry in `updateOne`'s release loop is removed, then restore it.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — Restore `properties.Create`'s return block indentation** (from CON-2, MNT-2)
Files: `core/database/properties/properties.go`.
Do: restore the leading whitespace of the `return Property{…}, []rowchange.Change{{…}}, nil` block to match `ae26c629`. `git diff -w ae26c629 -- core/database/properties/properties.go` must be unchanged by this commit.
Check: `git diff ae26c629 -- core/database/properties/properties.go` shows no whitespace-only hunk; scripts/check-pr.sh PASS.
