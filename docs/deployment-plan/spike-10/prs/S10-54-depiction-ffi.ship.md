# Ship log — S10-54 Depiction values over the FFI

**Plan:** [`S10-54-depiction-ffi.md`](S10-54-depiction-ffi.md) · **Branch:** `s10-54-depiction-ffi` · **PR:** #433
**State:** fixing 1

Base: `main` @ `233e87ad`. The branch was cut by the orchestrator and carries the plan refresh (`635bb9fd`, after S10-32 landed in #430; review log refresh rounds 1–2: Fits — Ready).

## Round 0 — implement
Head `d08a33c8`. Gate PASS (orchestrator run; Go and Xcode results cached). Report: deviations are extra test helpers (`likenessDraft`, `saveLikeness`, `checkSavedLikeness`, `likenessRegion2JSON`); Mac manual check not run. Stopped on: —.

## Round 1 — review
Base `233e87ad`, head `d08a33c8`. Verdicts: CON clean, MNT changes needed, UX clean, SEC clean (no findings), PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `depictionPaths` then `listedObservationProto(l, paths)` repeated at three sites; a nil map silently empties the path | Fix now → F-1 | verified: `observations.go:76-81`, `:412-418`, `source_read.go:70-76`. A wrapper keeps D-4's two functions, so no Decision changes |
| CON-1 | CON | minor → major | AC-2.1 read rows do not compare against the save response | Fix now → F-2 | re-rated: AC-2.1 is not evidenced as written; `checkSavedLikeness` checks only shape |
| CON-2 | CON | minor → major | T-2 does not check that exactly one object is added | Fix now → F-2 | re-rated: T-2 as planned ("gains exactly the returned path") is not what the test asserts |
| MNT-2 | MNT | minor | Two gates for "has a depiction" (`FromJoin` vs `listedFromSource`) | Fix now → F-1 (comment) | rides along: `source_read.go` is in F-1; equivalence holds by `NOT NULL` (`000052.sql`) |
| MNT-3 | MNT | nit | Stray blank line in `observationProto` | Fix now → F-1 | rides along; one line in a file F-1 touches |
| CON-3 | CON | nit | Landed note misses `mapDepictionInput` and `depiction` without `= nil` | Fix now → F-3 | one clause in a doc F-3 touches |
| UX-1 | UX | minor | `ingest.unsupported_image` copy has no next step | Rejected | the copy is the user's answer in the plan review log ("keep both", 2026-10-17); raised in the report for the user |
| UX-2 | UX | minor | An oversized crop shows import copy (`ingest.too_large`, 512 MB title) | Follow-up | S10-34: render crops at ≤ 2048 px long edge (what Go stores), or crop-specific copy; the plan's Out row covers import copy |
| PRF-1 | PRF | minor | Image decode runs under the catalog session lock | Follow-up | D-3 and S10-53 place staging inside `withProjectCatalog`; splitting the staging API is an S10-53 shape change |
| PRF-2 | PRF | minor | Create and append list the whole Citation to return written rows | Logged | bounded by Observations per Citation, indexed; same as `CreateCitedBridge` before this PR |

### Fix plan 1

**F-1 — One helper turns listed rows into response protos** (from MNT-1, MNT-2, MNT-3)
Files: `api/ffi/handlers/observations.go`, `api/ffi/handlers/source_read.go`. Skill: `add-ffi-handler`.
Failing test first: none (a refactor with no behavior change); the existing likeness rows of `TestUpdateObservation`, `TestGetCitation`, `TestListObservationsBySource`, `TestCreateCitationWithObservations`, `TestAddObservationsToCitation`, `TestCreateCitedBridge` stay green unedited.
Do: add `func listedProtos(c *database.Catalog, rows []observations.Listed) ([]*engine.Observation, error)` in `observations.go` beside `depictionPaths`: call `depictionPaths(c, rows)` once, then `listedObservationProto(l, paths)` for each row, in order. `listedWritten` and `listedProtosFromSource` end with a call to it (collect their rows, then `return listedProtos(c, rows)`); `UpdateObservation` calls `listedProtos(c, []observations.Listed{row})` and uses `[0]`. After this, `listedObservationProto` and `depictionPaths` have exactly one caller each (`listedProtos`); say so in `listedObservationProto`'s doc comment ("call through listedProtos"). In `listedFromSource`, add a one-line comment on the `len(o.DepictionID) > 0` gate: it matches `depictionvalues.FromJoin`'s non-nil result because `depiction_values.image_file_id` is `NOT NULL`. Remove the blank line before `ValueDepictionId` in `observationProto` and let gofmt realign.
Check: `rg -n "depictionPaths\(|listedObservationProto\(" api/ffi/handlers` shows only the definitions and `listedProtos`; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; scripts/check-pr.sh PASS.

**F-2 — The likeness read and create tests assert what AC-2.1 and T-2 say** (from CON-1, CON-2)
Files: `api/ffi/handlers/citations_test.go`.
Failing test first: not applicable (strengthens assertions on passing code); each new assertion must be seen to fail once by temporarily breaking it (e.g. compare against a wrong id), then restored — note it in the report.
Do: in `TestGetCitation` "returns a likeness" and `TestListObservationsBySource` "returns a likeness", keep the `saveLikeness` result and assert `got.GetValueDepictionId() == saved.Observations[0].GetValueDepictionId()` and `proto.Equal(got.GetDepiction(), saved.Observations[0].GetDepiction())`, as `TestUpdateObservation` "keeps a likeness on a polarity change" does. In `TestCreateCitationWithObservations` "saves a likeness and returns its image path", take `before := objectRelPaths(t, dir)` in `reqFn` after `likenessFixture` (closure variable), and in `after` assert the new `objectRelPaths` equals `before` plus exactly the returned path (sorted).
Check: those rows pass; scripts/check-pr.sh PASS.

**F-3 — Landed note lists the remaining differences** (from CON-3)
Files: `docs/deployment-plan/spike-10/prs/S10-54-depiction-ffi.md`.
Do: add to the Landed note: `GoStore` gained a private `mapDepictionInput` shared by the draft and update-request mappings; `CatalogObservationDraft.depiction` is declared without `= nil` (same behavior); responses go through one `listedProtos` helper (F-1).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-54-depiction-ffi.md` passes.
