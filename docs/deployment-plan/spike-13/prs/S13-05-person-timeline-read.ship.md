# Ship log — S13-05 Person timeline read

**Plan:** [`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md) · **Branch:** `s13-05-person-timeline-read` · **PR:** #680
**State:** ready

## Round 0 — implement
Head `bd726bd7`. Gate PASS (orchestrator run against `ed4f1b91`). Report: one test deviation — `TestListPersonEvents` asserts B2's participants as a set {Kid, Sam}; order is asserted in `TestPersonEvents/AC-1.3` (D-13, refs compare as strings). Recorded in the plan's Landed note.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX routed on `macos/App/**` (CatalogQueryKey, registry, stores; no strings or error codes). SEC+PRF routed on `api/ffi/**` and `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | minor | `rolesByNeighbor` role de-dup scans the neighbor's roles per link: links × roles on a catalog with duplicate associations | Fix now → F-1 | verified: `timeline.go:57-63` inner `bytes.Equal` loop over `out[i].roles`; code this PR adds; contradicts the plan's Edges "No dimension multiplies past the links read" |
| MNT-1 | MNT | minor | Year-date test builder copied again (`yearValue`, `year :=`), 11 hand-rolled copies across 9 test files | Fix now → F-2 (same-pass cleanup) | verified: `rg 'datevalue.KindPoint, Calendar: "gregorian", StartYear: &'`. The shared builder already exists, `datevaluetest.YMD` (used by valuecompare, subjectcompare, graphalign…), so no new helper; move every copy onto it |
| MNT-2, CON-3 | MNT, CON | minor | `hops.go` block comment says the participation rule three times; over-long rewrapped line | Fix now → F-3 | verified: `core/hops/hops.go:54-63` |
| CON-1, MNT-4 | CON, MNT | minor | `termsProto` is new, undocumented, away from `termProto`, and missing from the Landed note | Fix now → F-4 | verified: `api/ffi/handlers/person_detail.go:211`; `termProto` lives in `conclusion_summaries.go` |

UX: no findings (outside-my-lens note is CON-2/MNT-5, logged below).

Nits (logged): `CON-2 / MNT-5 · macos/App/Platform/FakeStore.swift:98 · seededPersonEvents lacks the doc line its sibling seededPersonFamilies has` · `MNT-3 · core/database/persondetail/timeline.go:89 · v.TypeIDsByKey() built twice; call once`

### Fix plan 1

**F-1 — De-duplicate roles per neighbor in O(1)** (from SEC-1)
Files: `core/database/persondetail/timeline.go`, `core/database/persondetail/timeline_drop_test.go` (or wherever `TestRolesByNeighbor` lives).
Failing test first: none can fail on behavior (output is unchanged); add a `TestRolesByNeighbor` case with one neighbor reached by many links carrying repeated and distinct role terms and assert each distinct role appears once, in first-seen order — this pins the contract the rewrite must keep.
Do: in `rolesByNeighbor`, keep a per-neighbor `map[string]struct{}` of term ids already seen (beside `index`, e.g. a `seen` field on `neighborRoles` or a parallel slice of maps), replacing the inner `bytes.Equal` scan. Mark a term seen only when `label` returned ok, matching today's behavior (an unlabeled term is retried on the next link — keep that exact semantics or, if simpler, mark seen either way and note it; output must be identical). Drop the `bytes` import if it becomes unused. While here, MNT-3 nit: call `v.TypeIDsByKey()` once at `timeline.go:89`.
Check: `go test -tags fts5 ./core/database/persondetail/` passes; scripts/check-pr.sh PASS.

**F-2 — Move every hand-rolled Gregorian year date onto `datevaluetest.YMD`** (from MNT-1; same-pass cleanup)
Files: `core/database/persondetail/timeline_test.go`, `core/database/persondetail/family_test.go`, `api/ffi/handlers/person_detail_test.go`, `api/ffi/handlers/conclusion_summaries_test.go`, `core/database/promotealign/promotealign_test.go`, `core/search/handles_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`, `core/database/promote/batch_test.go`.
Failing test first: none (test refactor; behavior unchanged).
Do: replace each `datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &y}` (and `&datevalue.Value{…}`) with `datevaluetest.YMD(y, 0, 0)` (dereference with `*` where a value, not a pointer, is needed). `core/search/handles_test.go:381` carries `StartMonth: &m` → `datevaluetest.YMD(y, m, 0)`. Delete the local helpers that become empty wrappers (`yearValue` in timeline_test.go, the `year :=` closures in family_test.go and person_detail_test.go, `pointYear` in conclusionsummaries_test.go, the helper in promote/batch_test.go) and call `datevaluetest.YMD` directly. Leave `core/valuecompare/datewindow_test.go:318` alone: it builds year 0, which `YMD` leaves unset by design. Find copies by body, not name: also check for multi-line literals of the same shape.
Check: `rg -n 'KindPoint, Calendar: "gregorian", StartYear: &' --type go` prints only `core/valuecompare/datewindow_test.go:318`; `go test -tags fts5 ./core/... ./api/...` passes; scripts/check-pr.sh PASS.

**F-3 — One sentence for the participation hops** (from MNT-2, CON-3)
Files: `core/hops/hops.go`.
Do: fold lines 54-63 into one statement of the difference, e.g. "A participation joins a person to an event; role stays off the association's identity and walks read it off the link. EventsOfSubject and SubjectsOfEvent keep only subject-role links; EventsOfParticipant and ParticipantsOfEvent keep every role." Drop the redundant inline comment; rewrap to the file's width.
Check: gofmt clean; scripts/check-pr.sh PASS.

**F-4 — `termsProto` beside `termProto`, documented, and in the Landed note** (from CON-1, MNT-4)
Files: `api/ffi/handlers/person_detail.go`, `api/ffi/handlers/conclusion_summaries.go`, the plan's Landed note.
Do: move `termsProto` next to `termProto` in `conclusion_summaries.go` with a one-line doc comment; add one line to the plan's Landed note naming it (a mapper the plan did not name; no behavior change). Rider (nit CON-2/MNT-5): give `FakeStore.seededPersonEvents` the same doc line as `seededPersonFamilies`.
Check: scripts/check-pr.sh PASS.

All four are minors: no reviewer goes back for round 2; the orchestrator reads the fix commits and marks each Resolved.

## Round 1 — fix
Head `079c6b3e`. Gate PASS. F-1 done (`827fc77a`, per-neighbor `seen` set; `TestRolesByNeighborManyLinks`), F-2 done (`ec0d2c02`; the grep prints only `core/valuecompare/datewindow_test.go:318`; `promote/batch_test.go`'s helper had no callers and was deleted; `pointYear`'s other callers in `conclusionsummaries` moved too), F-3 done (`c8836626`), F-4 done (`079c6b3e`, nit CON-2/MNT-5 rode along; MNT-3 rode along in F-1).
Orchestrator read of the fix commits: SEC-1 Resolved, MNT-1 Resolved, MNT-2/CON-3 Resolved, CON-1/MNT-4 Resolved.

## Round 2 — review
Not run: round 1 had no blocker or major.

## Finish
Merged `origin/main` @ `ae789572` (#679, docs only). Gate PASS against `ae789572`.

## Result
Gate PASS · CI green · open: none · follow-ups: none · to user: — · CON miss: none
