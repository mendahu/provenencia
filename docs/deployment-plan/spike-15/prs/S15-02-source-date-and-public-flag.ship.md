# Ship log — S15-02 Sources carry a date and a not-public flag

**Plan:** [`S15-02-source-date-and-public-flag.md`](S15-02-source-date-and-public-flag.md) · **Branch:** `s15-02-source-date-and-public-flag` · **PR:** #762
**State:** stopped — CI Go budget red twice, to user

## Round 0 — implement
Head `940f719a` (PR #762, draft). Gate PASS (orchestrator re-run on `940f719a` against `462699f1`).
Report deviations: `TestProductAliasTablesPinCacheVersion` pin 23→24 with D-17's `CacheVersion` bump; T-9's test lives in `deleteimpact/release_audit_test.go` (external test package, import cycle); version 0.4.4→0.5.0 (main moved to 0.4.4 after planning); extra assertions (unknown Source, short id, one date row per Source). Implementer saw one flaky xcode-test run (`WorkspaceDestinationHostTests` canvas tests, "Expectation failed: loaded") on a docs-only commit; re-run passed. Manual check step 1 not done (user's dogfood project).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX dispatched on `macos/App/Platform/Generated/engine.pb.swift` (generated wire code only; no strings or view changes) — cheap to run. SEC+PRF: migration, `api/ffi/**`, `core/database/**`, `core/autoreconcile`, `core/valuecompare`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | "Blank kind means no date" written in two handlers | Fix now → F-1 | verified: `handlers/observations.go:296` and `handlers/sources.go:369` hold the same test around `valuecodec.DateFromProto` |
| SEC-1 | SEC | minor | Dangling `sources.date_value_id` makes a listed Source unopenable | Fix now → F-2 | verified: `sources.go` `getOne` returns `LookupTx`'s `sql.ErrNoRows`; `List` treats the same miss as undated |
| CON-4 (from MNT, Outside my lens) | CON | minor | `ReleaseReplaced` guard loosened from `len != 16` to `len == 0` | Fix now → F-3 | verified: `release.go:118` vs base `observations/update.go:200`; a malformed stored id now reaches `rowMap`/`LookupTx` and fails the write instead of a no-op |
| MNT-2 | MNT | minor | Each owner hand-writes the lookup-then-RowMap closure | Fix now → F-4 | same-pass cleanup: this PR adds the second date copy (`sources.go:296`, `observations/update.go:95`) |
| MNT-3 | MNT | minor | Source write check-and-load preamble in five places | Fix now → F-5 | same-pass cleanup: this PR adds copies four and five (`SetDate`, `SetPublic`) |
| MNT-4 | MNT | minor | `PutTx` normalizes/validates, then `InsertTx`/`UpdateTx` do it again | Fix now → F-6 | verified: `datevalues.go:172-177` then `InsertTx`'s own `normalize`/`Validate`; same in `namevalues`; code this PR adds |
| CON-1 | CON | minor | `DateWindow` comment still says sort keys read it | Fix now → F-7 | verified: `core/valuecompare/datewindow.go:301`; stale comment in a file family this PR changes; contradicts "One date order" |
| SEC-2 | SEC | minor | Source and Observation could share one date row in a crafted catalog | Follow-up | adversarial-only: `sources_date_value_id_uidx` + `observations_value_date_id_uidx` + every write minting rows through `PutTx` keep the app from creating it; a refused clear/delete is safe (FK `NO ACTION`); the in-place edit is audited on both scopes (D-9). The right fix is one owner check in `datevalues.PutTx` over every owned FK into `date_values`, not a Sources-only guard. Ask the user where it goes |
| UX-1 | UX | minor | Write to a Source deleted elsewhere says "Invalid source." | Later PR (S15-03) | written into S15-03's map scope entry ("From S15-02's review"), commit `972ba6cf` |

Nits (logged): `UX-2 · Localizable.xcstrings:3250 · "Invalid date value." could name the field in S15-03 (also put in S15-03's map entry)` · `MNT-5 · namevalues.go:72,110 · leftover parts := v.Parts; use namevalue.Validate(v) (folded into F-6)` · `MNT-6 · sources.go:363 · boolInt duplicates propertyterms.directedBit; move to database at a third copy` · `CON-2 · plan Landed note · list the add-match-rule SKILL edit (folded into F-7)` · `CON-3 · handlers/sources_test.go · use cmp.Compare in the date-sort comparator (folded into F-7)`

### Fix plan 1

**F-1 — One decoder for an optional wire date** (from MNT-1)
Files: `core/valuecodec/valuecodec.go`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/observations.go`.
Failing test first: `valuecodec` `TestOptionalDateFromProto` — nil input → nil; kind `"  "` → nil; kind `"point"` year 1900 → non-nil with year 1900.
Do: add `OptionalDateFromProto(d *engine.DateValueInput) *datevalue.Value` (nil when `d` is nil or its kind is blank after trim, else `DateFromProto`). Move both handlers onto it; drop the `strings` import where unused.
Check: `grep -rn "strings.TrimSpace(.*GetKind())" api core` prints nothing; `TestSetSourceDate` and the observations handler tests pass; scripts/check-pr.sh PASS.

**F-2 — A Source whose date row is missing reads as undated** (from SEC-1)
Files: `core/database/sources/sources.go`, `core/database/sources/sources_test.go`.
Failing test first: `TestSourceDate/dangling_date_row_reads_undated` — with `PRAGMA foreign_keys=OFF` on a test connection, point a Source's `date_value_id` at 16 random bytes; `Get` returns the Source with `Date == nil` and no error; `SetDate` with a date then succeeds (writes a fresh row) and `SetDate` with nil clears the column.
Do: in `getOne`, treat `sql.ErrNoRows` from `datevalues.LookupTx` as undated (leave `Date` nil, keep `DateValueID`), matching `List`. In `SetDate`, when the stored id has no row, write through `PutTx` with a nil `prevID` and skip `ReleaseReplaced` for the missing row (or have F-3's guard and the rowMap's miss make it a no-op — whichever keeps one rule; say which in the report).
Check: the new case passes; `TestSourceDate`, `TestSourceDelete` pass; scripts/check-pr.sh PASS.

**F-3 — `ReleaseReplaced` ignores a malformed previous id again** (from CON-4)
Files: `core/database/deleteimpact/release.go`, `core/database/deleteimpact/release_audit_test.go`.
Failing test first: `TestReleaseReplacedAudits/malformed_prev_id_is_a_no_op` — `prevID` of 3 bytes returns `(nil, nil)` and does not call `rowMap`.
Do: change the guard to `len(prevID) != 16 || bytes.Equal(prevID, newID)`, as the base `releaseReplacedValue` had.
Check: the case passes; scripts/check-pr.sh PASS.

**F-4 — Each value package exports its release row-map once** (from MNT-2; same-pass cleanup)
Files: `core/database/datevalues/datevalues.go`, `core/database/namevalues/namevalues.go`, `core/database/depictionvalues/depictionvalues.go`, `core/database/sources/sources.go`, `core/database/observations/update.go`, `core/database/observations/depiction.go`, and any other `ReleaseReplaced` caller.
Do: add `RowMapTx(tx *sql.Tx, id []byte) func() (map[string]any, error)` to each of the three packages (lookup then `RowMap`). Move every `ReleaseReplaced` caller onto it.
Check: `grep -rn -A4 "ReleaseReplaced(" core --include=*.go | grep -E "func\(\) \(map\[string\]any, error\)"` prints only the three `RowMapTx` definitions (and test fakes in `_test.go`); scripts/check-pr.sh PASS.

**F-5 — One load-for-write helper for Source writes** (from MNT-3; same-pass cleanup)
Files: `core/database/sources/sources.go`.
Do: add private `loadForWrite(tx *database.Tx, userID, id []byte) (Source, error)` (tx/id-length check, `RequireUserID`, `getTx`, `sql.ErrNoRows` → `ErrInvalid`). Use it in `SetCover`, `SetDate`, `SetPublic`, `Update`, `Delete`. Keep each function's existing error results unchanged.
Check: `grep -c "errors.Is(err, sql.ErrNoRows)" core/database/sources/sources.go` counts only `loadForWrite` plus any non-write uses (name them in the report); all `sources` tests pass unchanged; scripts/check-pr.sh PASS.

**F-6 — Normalize and validate once per value write** (from MNT-4, MNT-5)
Files: `core/database/datevalues/datevalues.go`, `core/database/namevalues/namevalues.go`.
Do: in each package split the stored write into private `insert`/`update` taking an already-normalized, validated value. `InsertTx`, `UpdateTx` and `PutTx` each normalize and validate once, then call the private step. In `namevalues`, validate `v` directly (drop the leftover `parts := v.Parts`).
Check: `datevalues` and `namevalues` `TestPutTx` and existing tests pass unchanged; scripts/check-pr.sh PASS.

**F-7 — Doc and test tidy-ups** (from CON-1, CON-2, CON-3)
Files: `core/valuecompare/datewindow.go`, `api/ffi/handlers/sources_test.go`, the plan's Landed note.
Do: drop "Sort keys" from the `DateWindow` comment (where a date sorts is `DateKeyOf`). Use `cmp.Compare` in the date-sort comparator. Add the `add-match-rule` SKILL edit, and F-1–F-6, to the plan's Landed note.
Check: `grep -n "Sort keys" core/valuecompare/datewindow.go` prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b08bfaab`. Gate PASS (orchestrator re-run). F-1 – F-7 done, each with its test red first where named. F-2 keeps one rule: `held` is the stored id only when its row exists; a new date inserts a fresh row and repoints; a clear passes `held` (nil) to `ReleaseReplaced`. Report deviations: F-1 grep still prints the definition inside `OptionalDateFromProto`; F-4 grep also matches the `ReleaseReplaced` signature and F-3's test fake (`observations/depiction.go` had no caller); F-5 leaves 5 non-write `sql.ErrNoRows` uses (`getOne`, `requireType`, `CoverArtifactFile`, `requireRasterCoverArtifact`, `loadForWrite`); `SetCover` now validates the cover mode after the load (same `ErrInvalid`).

Minors checked by the orchestrator (no round-2 reviewer for CON or SEC — neither lens had a Fix-now blocker or major):
- SEC-1 **Resolved** — `getOne` treats `sql.ErrNoRows` as undated (`3243c51b`); `TestSourceDate/dangling_date_row_reads_undated`.
- CON-4 **Resolved** — guard is `len(prevID) != 16` again (`f7a10d52`), doc comment updated.
- CON-1 **Resolved** — `datewindow.go:301` no longer names sort keys; points at `DateKeyOf`.

## Round 2 — review
Reviewers: MNT (MNT-1 was the round's only Fix-now major). Scope `940f719a...b08bfaab`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) blank kind in two handlers | Resolved | `valuecodec.OptionalDateFromProto` is the one decoder; both handlers call it; grep finds only its definition |

No new findings. Round-1 minors MNT-2 – MNT-5 confirmed done by the round-2 reviewer (`RowMapTx` ×3, `loadForWrite` ×5, private `insert`/`update`). Verdict: Clean.

## Result
`origin/main` unchanged since the branch was cut (`462699f1`); no merge needed. Gate PASS on `b08bfaab` · CI: macOS green; Go red twice on `promotealign` `TestLargeSourceBudgets/compose_every_summary` (568 ms, then 510 ms; budget 500 ms) — local head 208–233 ms vs base 217–333 ms, alone and under the full suite; to user · open: none · follow-ups: SEC-2 → [`docs/ideas/value-row-owner-check.md`](../../../ideas/value-row-owner-check.md) (user, 2026-11-09) · later PR: UX-1, UX-2 → S15-03 map entry · to user: —
