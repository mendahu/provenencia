# Ship log — S10-32 Depiction value and `likeness` on Person

**Plan:** [`S10-32-depiction-value-type.md`](S10-32-depiction-value-type.md) · **Branch:** `s10-32-depiction-value` · **PR:** #430
**State:** ready

The plan refresh (`36ef0065`) is the branch's first commit, on the user's call (2026-10-17): no separate docs PR. Base `ae26c629`.

## Round 0 — implement
Head `0532a08b`. Gate PASS (orchestrator run, against `ae26c629`). Report deviations: STEP-6's `go list -deps ./core/database/...` check was wrong in the plan (`sources` and `sourcemetadata` already import `core/writes` on `main`); `go list -deps ./core/database/observations` shows neither `ingest` nor `writes`. Depiction insert body is `resolveDepiction` in `depiction.go`; `releaseReplacedValue` runs from one table loop in `updateOne`; STEP-6 and STEP-7 landed together; extra T-6/T-7 rows and test helpers. Mac manual check not run.

## Round 1 — review
Five lenses on `ae26c629...0532a08b`. Verdicts: CON changes needed; MNT, UX, SEC, PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | T-7's date-release clause has no test | Fix now → F-1 | Plain: no test replaces a date value on update (`rg 'FROM date_values'` hits only the delete row); reviewer ran it via `-overlay`, behavior correct |
| CON-2, MNT-2 | CON, MNT | minor | Whitespace-only re-indent of `properties.Create` return block | Fix now → F-2 → Rejected | F-2 not reproduced: the re-indent is gofmt's output; `gofmt -l` flags `ae26c629`'s own `properties.go`, and the gate gofmts every changed file |
| CON-3 | CON | minor | Manual check §10.1 not run | Follow-up → user's manual check | The user runs the manual check before merge (ship-pr step 8) |
| MNT-1 | MNT | minor → blocker (user) | `ValidateDepictionRegion` decodes twice via a second decoder | Fix now → F-3 (user, 2026-10-17) | `core/locator` split into `decode` + `validateDoc`; pairs with SEC-1's point cap, before S10-54 |
| MNT-3 | MNT | minor → blocker (user) | `resolvedValue` carries value changes three ways | Fix now → F-4 (user, 2026-10-17; overrides D-8) | Contradicts D-8's named fields; log with M-8 (declared value-column list) |
| MNT-4 | MNT | nit | `sameStrings` duplicates `slices.Equal` | Logged | |
| MNT-5 | MNT | nit | Region trimmed twice; comment names `Insert` | Logged | |
| UX-1 | UX | minor | VoiceOver reads Likeness value type as *Text* | Follow-up → S10-54 | Plan Risks row accepts it; S10-54 adds the label |
| UX-2 | UX | nit | "More than one is allowed." vs "More than one value is allowed." | Logged | D-4 fixes the text |
| SEC-1 | SEC | minor | No cap on region point count; O(n²) self-intersection | Rejected (user, 2026-10-17: descoped) | Same gap exists for Citation locators; unreachable until S10-54 |
| SEC-2 | SEC | minor | `StageDepictions` holds every staged PNG; no per-call cap | Follow-up → S10-54 | Handler-level cap when the FFI field lands |
| SEC-3 | SEC | minor → blocker (user) | `Listed.Depiction.RegionJSON` unvalidated on read | Fix now → F-5 (user, 2026-10-17) | Validate and clamp before drawing or cropping |
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

## Round 1 — fix
Head `b41f9b06`. Gate PASS (orchestrator run). F-1 done (`2728a9b2`): `TestObservations/date_replaced_by_another_Property_releases_its_value` passes and fails with `deleteimpact.invalid` when the `observations.value_date_id` release entry is removed. F-2 not applied: restoring `main`'s indentation (`0968541f`) failed the gofmt gate, and `main`'s file fails `gofmt -l` itself; reverted in `b41f9b06`. Dropped CON-2 / MNT-2 as Rejected.

## Round 2 — review
CON and MNT (the lenses with Fix-now findings) on `0532a08b...b41f9b06`. Both clean, no new findings. CON-1 resolved: the new test fails (`replaced date value left behind: 1 rows`) when the date release entry is removed. CON-2 / MNT-2 rejection confirmed (`gofmt -l` flags `main`'s `properties.go`).

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-1 + SEC-1 (locator decode split and point cap, before S10-54), SEC-2 and UX-1 (S10-54), SEC-3 (S10-34, S10-36), MNT-3 (with M-8) · to user: manual check §10.1 (CON-3)

## User decisions (2026-10-17)

After the report, the user re-rated four round-1 findings and asked for stricter blocking criteria in the review skills, on this PR.

| ID | Was | Now | User's reason |
| --- | --- | --- | --- |
| MNT-1 | minor, Follow-up | blocker, Fix now → F-3 | A second decoder in `core/locator` is a blocker |
| MNT-3 | minor, Follow-up (contradicts D-8) | blocker, Fix now → F-4 | One way to carry value changes; overrides D-8's three named fields |
| SEC-3 | minor, Follow-up → S10-34/36 | blocker, Fix now → F-5 | A known later caller is not a reason to defer: "If we know the PR is coming, then we should just fix it now" |
| SEC-1 | minor, Follow-up | Rejected (descoped) | Hand-drawn regions will not approach a point count that matters |

Skills changed in `.agents/skills` (commit "Rate review findings by the code as merged…"): `review-pr` "Rate the code as merged"; `maintainability.md` Q8, Q14, severity table, Known traps; `security.md` step 4, "Reads that pass untrusted values through", severity table; `ship-pr` triage "Re-rate before you decide"; `review-plan` blocker row.

### Fix plan 2

**F-3 — One locator decoder** (from MNT-1)
Files: `core/locator/validate.go`, `core/locator/depiction.go`, `core/locator/depiction_test.go`.
Failing test first: none (refactor); `TestValidateDepictionRegion` and the existing `validate_test.go` suite must pass unedited before and after.
Do: split `Validate` into `decode(locatorJSON string) (document, error)` (trim, empty check, unmarshal, version and selector-count checks — everything `Validate` does before its selector loop) and `validateDoc(doc document) error` (the selector loop). `Validate` becomes `decode` then `validateDoc`. `ValidateDepictionRegion` calls `decode` once per document, then `validateDoc` on each, then its own region and page rules. Delete `parse`. `json.Unmarshal` appears once in `core/locator` outside tests.
Check: `CGO_ENABLED=1 go test ./core/locator/...` passes; `grep -n "json.Unmarshal" core/locator/*.go | grep -v _test` prints one line.

**F-4 — One list of value changes in `resolvedValue`** (from MNT-3; overrides D-8, user 2026-10-17)
Files: `core/database/observations/observations.go`, `core/database/observations/update.go`, `core/database/observations/depiction.go`.
Failing test first: none (refactor); every `TestObservations` row must pass unedited, including the change-order assertions.
Do: replace `resolvedValue.DateChange`, `NameChange`, and `DepictionChanges` with one `ValueChanges []rowchange.Change`. The date and name cases append their one change to it; `resolveDepiction` sets the file create change (if any) then the `depiction_value` create change, as today. `insertOne` and `updateOne` each do `changes = append(changes, resolved.ValueChanges...)` once, where the three appends are today, so the change order is unchanged. Delete the old fields and their nil-checked appends.
Check: `grep -rn "DateChange\|NameChange\|DepictionChanges" core` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**F-5 — Validate a stored depiction region when it is read** (from SEC-3)
Files: `core/database/depictionvalues/depictionvalues.go`, `core/database/depictionvalues/depictionvalues_test.go`, `core/database/observations/observations.go`, `core/database/observations/observations_test.go`, `core/locator/depiction.go`, `core/locator/depiction_test.go`. In the `depictionvalues` package comment, also fix MNT-5's `Insert` → `InsertTx`.
Failing test first: `TestObservations/likeness_read_drops_a_stored_region_that_is_not_valid` — save a likeness as in T-5, then with raw SQL on `c.DB()` set its `depiction_values.region_json` to each of: `not json`; a region locator whose points include `{"x":5,"y":5}`; a self-intersecting four-point region; a document with no `region` selector. For each, `ListByCitation` and `ListBySubject` (on the Person) return the Observation with `ValueDepictionID` set, `Depiction` non-nil, `Depiction.ImageFileID` equal to the stored File id, and `Depiction.RegionJSON == ""`. A valid stored region is returned unchanged (existing T-5 assertion).
Do: add `func Readable(v Value) Value` *(new)* in `depictionvalues`: it returns `v` unchanged when `locator.Validate(v.RegionJSON)` passes and the document has exactly one `region` selector; otherwise it returns `v` with `RegionJSON` set to `""`. Add the one-region-selector check to `core/locator` as an exported `ValidateRegionDocument(regionJSON string) error` *(new)* in `depiction.go`, built on F-3's `decode` / `validateDoc`, and have `ValidateDepictionRegion` call it for the region half so the region rules live once; `Readable` calls `ValidateRegionDocument`. `scanListed` passes every hydrated `Depiction` through `depictionvalues.Readable`. The image id is kept: the image is Go's own stored PNG, safe to show without the region. `LookupTx` (used for audit row maps) is unchanged: audit records what is stored. Document on `Listed.Depiction` that an empty `RegionJSON` means the stored region failed validation, and that every later reader that returns a region (S10-54's source-graph hydration) goes through `Readable`.
Check: the test passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/... ./core/locator/...` passes; scripts/check-pr.sh PASS.

Also in this round (docs, the implementer): the plan's **Landed** note records F-3 to F-5 and that F-4 overrides D-8 by the user's call; the PR follow-ups no longer list MNT-1, MNT-3, SEC-3, or SEC-1.

## Round 2 — fix (user's re-rating)
Head `ab848b35`. Gate PASS (orchestrator run). F-3 done (`a1a96ed6`): one `json.Unmarshal` in `core/locator`; `parse` deleted. F-4 done (`dd129ae8`): `resolvedValue.ValueChanges`; old fields gone; change-order tests pass unedited. F-5 done (`ac70707d`): `depictionvalues.Readable` on every `scanListed` depiction, `locator.ValidateRegionDocument`; the new test fails on all four bad regions with `Readable` disabled. No review round 3: the user decides.

## Result (after user re-rating)
Gate PASS · open: none · follow-ups: SEC-2, UX-1 (S10-54) · rejected by user: SEC-1 · to user: manual check §10.1 (CON-3); whether to run a review round on the round-2 fix commits
