# Ship log — S16-30 Provenance records compose from cached records

**Plan:** [`S16-30-provenance-from-records.md`](S16-30-provenance-from-records.md) · **Branch:** `s16-30-provenance-from-records` · **PR:** #812
**State:** ready

## Round 0 — implement
Head `fbc8e7bd9`. PR #812 (draft). Gate PASS (orchestrator run against `d1d13478e`). Report deviations:
- `TestGetConclusionDetailCarriesDerivedRows`' fixture lifted into `derivedRowsFixture` for T-6; assertions unchanged.
- T-5's new cases live in `TestFieldsOfOnly` beside `TestFieldsOf`.
- One earlier gate run failed on `TestLargeSourceBudgets/warm_at_open` (2.56 s vs 2.5 s, parallel suite); 0.70 s alone; reran green, no check changed.

Figures (Apple M1, base → after): `TestDetailBudgets` cold, 80 graphs 0.90 ms → 15.6 ms (budget 1 s); warm 0.65 ms → 0.11 ms; `BenchmarkGraphRecords` 669 ms / 17.15 MB → 736 ms / 22.38 MB; `graphRecords` at 16,000 persons 528,960 B → 695,544 B.

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, `Localizable.xcstrings`, `core/apperr`, or new error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT-2 | CON, MNT | major | T-7's shared `warmReadTables` dropped `depiction_values`; the test's doc comment now sits on the var | Fix now → F-1 | verified: `conclusion_detail_test.go:301-309` at head has no `depiction_values`; base list had it |
| CON-2 | CON | major | T-6's fixture has no Person name record, so the sheet's own records are never composed end to end | Fix now → F-2 | reviewer's overlay probe logged `fields: 0`; `connectFixture` adds only a subject label |
| SEC-1 | SEC (+PRF) | major | `loadGraphObservations` parses each Citation's locator and copies its words per Observation: locator bytes × Observations | Fix now → F-3 | verified: `load.go:635-646` scans `CitationLabel` and calls `locator.Page` per row; reviewer timed 2.06 ms per 64 KB parse. One parse per Citation is simpler than a cap and also cuts real-data growth (+31% `graphRecords`) |
| SEC-1 (depiction part) | SEC | major | `depictionvalues.FromJoin` → `depictionvalue.Readable` parses the same locator per depiction Observation | Follow-up | predates this PR, which neither adds nor worsens it; routing it through the per-Citation entry changes `FromJoin`'s contract |
| MNT-1 | MNT | major | Fourth hand copy of the warm-budget timing block | Fix now → F-4 | same-pass cleanup: `git grep "databasetest.Fastest(5"` lists four files, the new `detail_budget_test.go` among them |
| MNT-3, CON-4 | MNT, CON | minor | Test comment still names "record shells" | Fix now → F-5 | `conclusiondetails_test.go:579` |
| CON-3 | CON | minor | Landed note and PR body say "table list widened" and no deviations | Fix now → F-6 | true after F-1/F-2; refresh with the round's figures |

Nits (logged): `MNT-4 · conclusiondetails.go PersonDerivedRows · build only in the first loop` · `MNT-5 · conclusiondetails.go fieldsOf · wanted and rows are parallel slices` · `MNT-6 · conclusiondetails.go newRecordReader · test len(SubjectID) before lookup` · `CON-5 · graphfill/records_test.go fills · unused id parameter` (F-5 takes CON-5 and MNT-6 along; they are in files F-3/F-5 touch)

Outside-lens notes: CON → PRF "cold read 0.90 → 15.6 ms at 80 graphs": expected by the plan's Risks row 1, within the 1 s budget; no finding. CON → UX "manual check not run": the user runs §10.

### Fix plan 1

**F-1 — T-7 checks `depiction_values` again** (from CON-1, MNT-2)
Files: `api/ffi/handlers/conclusion_detail_test.go`.
Failing test first: none (coverage restore); confirm the test still passes and that removing `depiction_values` from its list is the only difference from base.
Do: `TestGetConclusionDetailWarmReadsNoReconcilerOrGallerySQL` loops over `append(slices.Clone(warmReadTables), "depiction_values")`. Give `warmReadTables` only its own comment, and move the old "A warm open composes the Why and the gallery …" comment back onto the test function, below the var.
Check: `go test ./api/ffi/handlers -run TestGetConclusionDetailWarm` PASS; `rg -n depiction_values api/ffi/handlers/conclusion_detail_test.go` prints the loop line.

**F-2 — T-6's Person carries a cited name record** (from CON-2)
Files: `api/ffi/handlers/conclusion_detail_test.go` (`derivedRowsFixture` or T-6 alone).
Failing test first: add to `TestGetConclusionDetailWarmReadsNoRecordSQL` an assertion that the first response has a `name` field with at least one outcome; it fails on the current fixture.
Do: before promoting, cite a name Observation on `personID` with `CreateCitationWithObservations`, `propertyIDByKey(t, dir, "name")` and a name draft, as the same fixture already does for the birth event. Keep `TestGetConclusionDetailCarriesDerivedRows`' assertions unchanged.
Check: T-6 passes with the name assertion and the same-bytes and no-table-read checks; scripts/check-pr.sh PASS.

**F-3 — Parse each Citation's words once per graph load** (from SEC-1)
Files: `core/database/graphfill/load.go`, `core/database/graphfill/records_test.go` (or a new `load_test.go`).
Failing test first: `TestLoadParsesEachCitationOnce` — one Citation with a locator near `locator.MaxDocumentBytes` (one page selector plus filler selectors `validateDoc` keeps) and N Observations; assert every Observation of that Citation shares one backing `PageLabel` / `CitationLabel` string (`unsafe.StringData` equal) and one `ArtifactID` slice (`&ArtifactID[0]` equal). Red today: each row has its own copy.
Do: in `loadGraphObservations`, keep a `map[string]citationWords` keyed by Citation id holding `ArtifactID`, `CitationLabel`, `ArtifactPage`, `PageLabel`, `TranscriptionUncertain`. On first sight of a Citation, clone `ArtifactID` and call `locator.Page` once; every later row of that Citation takes the stored values (drop the row's scanned copies). Leave `depictionvalues.FromJoin` as it is (its parse is a Follow-up). Re-run `BenchmarkGraphRecords` and `BenchmarkCacheSizes` and report build time, heap and `graphRecords` bytes against the round-0 figures.
Check: the new test passes; `TestGraphRecordsCarryCitationWords` (loaded and cold fills `DeepEqual`) passes; `PROVENENCIA_GRAPH_VERIFY=1 go test ./core/database/...` PASS; scripts/check-pr.sh PASS.

**F-4 — One warm-budget helper** (from MNT-1; same-pass cleanup)
Files: `core/database/databasetest/budget.go`, `core/database/persondetail/detail_budget_test.go`, `core/database/persondetail/claims_budget_test.go`, `core/database/persondetail/recordlinks_perf_test.go`, `core/database/promotealign/perf_test.go`.
Failing test first: none (refactor; behavior is the timing rule itself).
Do: add `databasetest.Warm(read func()) time.Duration`: one warm-up call, then `Fastest(5, func() time.Duration { runtime.GC(); return Since(read) })`. Move the promotealign comment's general part (why each run collects first; five runs for CI noise) onto `Warm`; keep promotealign's figure and idea link at its call site. Move all four callers onto it.
Check: `grep -rn "databasetest.Fastest(5" core api` prints nothing; `grep -rn "databasetest.Warm(" core` prints four call sites; the four budget tests pass; scripts/check-pr.sh PASS.

**F-5 — Stale term and nits** (from MNT-3, CON-4, CON-5, MNT-6)
Files: `core/database/conclusiondetails/conclusiondetails_test.go:579`, `core/database/graphfill/records_test.go`, `core/database/conclusiondetails/conclusiondetails.go`.
Do: "reach the record shells" → "reach the record through its Source record"; drop `fills`' unused `id` parameter; in `newRecordReader` test `len(row.SubjectID) > 0` before the map lookup.
Check: `git grep -n "record shells" -- '*.go'` prints nothing; scripts/check-pr.sh PASS.

**F-6 — Landed note and PR body** (from CON-3)
Files: `docs/deployment-plan/spike-16/prs/S16-30-provenance-from-records.md` (Landed note), PR #812 body.
Do: after F-1–F-5, state T-7's list as widened (AC-1.1's tables plus `depiction_values`), T-6's fixture with a cited name, the per-Citation parse (F-3) and the shared warm helper (F-4, superseding nothing in the plan), and the refreshed figures from F-3.
Check: scripts/check-pr.sh PASS (plan-docs).

## Round 1 — fix
Head `6b8669ebb`. Gate PASS (orchestrator). F-1 to F-6 done. F-2's name assertion and F-3's `TestLoadParsesEachCitationOnce` ran red first. Figures (round 0 → round 1): `BenchmarkGraphRecords` 736 ms / 22.38 MB → 743 ms / 22.00 MB; `graphRecords` at 16,000 persons 695,544 B → 691,464 B (the synthetic fixture has few Observations per Citation); `TestDetailBudgets` cold 15.6 → 13.3 ms, warm 0.11 → 0.10 ms.
Minors checked by the orchestrator: F-5 (MNT-3, CON-4, CON-5, MNT-6) **Resolved**: `git grep "record shells" -- '*.go'` empty; `fills` has no `id`; `newRecordReader` tests the Subject first. F-6 (CON-3) **Resolved**: Landed note and PR body name the widened T-7 list, the cited name, F-3 and F-4.

## Round 2 — review
Reviewers: CON, MNT, SEC+PRF (each had a Fix-now major). Scope: `fbc8e7bd9...6b8669ebb`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, CON-2 | CON | — | Round-1 majors | Resolved | T-7 loops over `warmReadTables` + `depiction_values` (`conclusion_detail_test.go:342`); `derivedRowsFixture` cites "Ada Lovelace" and T-6 asserts a `name` outcome |
| MNT-1, MNT-2 | MNT | — | Round-1 majors | Resolved | four `databasetest.Warm(` call sites, no `Fastest(5` outside `Warm`; T-7 comment back on the test |
| SEC-1 | SEC+PRF | — | Round-1 major | Resolved | `load.go:185-223` keeps `citationWords` per Citation; `TestLoadParsesEachCitationOnce` proves shared storage |

No new blockers from any reviewer.

Nits (logged): `SEC+PRF · graphfill/load.go loadGraphObservations · citationLocator still scanned into sql.NullString per row (short-lived copy); sql.RawBytes would drop it`

## Result
Gate PASS (head `6b8669ebb`, `main` unchanged at `d1d13478e`) · CI pending · open: none · follow-ups: SEC-1 depiction part (`depictionvalues.FromJoin` parses the Citation locator once per depiction Observation) · to user: —
